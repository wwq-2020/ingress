package server

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"sort"
	"sync"
	"time"

	v1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// Server Server
type Server struct {
	informerFactory informers.SharedInformerFactory
	defaultURL      *url.URL
	sync.Mutex
	doneCh        chan struct{}
	host2Backends map[string][]*backend
	httpServer    *http.Server
}

// New New
func New(clientSet kubernetes.Interface) *Server {
	informerFactory := informers.NewSharedInformerFactory(clientSet, time.Minute)

	ingresses := informerFactory.Extensions().V1beta1().Ingresses()

	server := &Server{
		informerFactory: informerFactory,
		doneCh:          make(chan struct{}),
		host2Backends:   make(map[string][]*backend),
	}
	httpServer := &http.Server{
		Addr:    ":8001",
		Handler: server,
	}
	server.httpServer = httpServer

	ingresses.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    server.onIngressAdd,
			UpdateFunc: server.onIngressUpdate,
			DeleteFunc: server.onIngressDelete,
		})

	informerFactory.Start(server.doneCh)

	return server
}

// Start Start
func (s *Server) Start() {
	s.informerFactory.WaitForCacheSync(s.doneCh)
	if err := s.httpServer.ListenAndServe(); err != nil {
		log.Fatalf("failed to ListenAndServe:%#v", err)
	}
}

// ServeHTTP ServeHTTP
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Lock()
	defer s.Unlock()
	var targetURL *url.URL
	backends, exist := s.host2Backends[r.Host]
	if exist {
		if s.defaultURL == nil {
			goto notfound
		}
		targetURL = s.defaultURL
		goto proxy
	}
	for _, each := range backends {
		if each.match(r.URL.Path) {
			targetURL = each.url
			goto proxy
		}
	}
notfound:
	w.WriteHeader(http.StatusNotFound)
	return

proxy:
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.ServeHTTP(w, r)
}

// Stop Stop
func (s *Server) Stop() {
	close(s.doneCh)
}

func (s *Server) onIngressAdd(obj interface{}) {
	s.Lock()
	defer s.Unlock()
	ingress, ok := obj.(*v1beta1.Ingress)
	if !ok {
		return
	}
	spec := ingress.Spec
	if spec.Backend != nil {
		s.defaultURL = &url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%s", spec.Backend.ServiceName, spec.Backend.ServicePort.String()),
		}
	}
	for _, rule := range spec.Rules {
		if rule.HTTP == nil {
			continue
		}
		paths := rule.HTTP.Paths
		sort.Slice(paths, func(i, j int) bool {
			return paths[i].Path > paths[j].Path
		})
		for _, path := range rule.HTTP.Paths {
			pathReg, err := regexp.Compile(path.Path)
			if err != nil {
				continue
			}
			backendObj := path.Backend
			s.host2Backends[rule.Host] = append(s.host2Backends[rule.Host], &backend{
				path: pathReg,
				url: &url.URL{
					Scheme: "http",
					Host:   fmt.Sprintf("%s:%s", backendObj.ServiceName, backendObj.ServicePort.String()),
				},
			})
		}
	}
}

func (s *Server) onIngressUpdate(oldObj, newObj interface{}) {
	s.Lock()
	defer s.Unlock()
	oldIngress, ok := oldObj.(*v1beta1.Ingress)
	if !ok {
		return
	}
	newIngress, ok := newObj.(*v1beta1.Ingress)
	if !ok {
		return
	}

	spec := oldIngress.Spec
	if spec.Backend != nil {
		s.defaultURL = nil
	}
	for _, rule := range spec.Rules {
		if rule.HTTP == nil {
			continue
		}
		delete(s.host2Backends, rule.Host)
	}

	spec = newIngress.Spec
	if spec.Backend != nil {
		s.defaultURL = &url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%s", spec.Backend.ServiceName, spec.Backend.ServicePort.String()),
		}
	}
	for _, rule := range spec.Rules {
		if rule.HTTP == nil {
			continue
		}
		paths := rule.HTTP.Paths
		sort.Slice(paths, func(i, j int) bool {
			return paths[i].Path > paths[j].Path
		})
		for _, path := range rule.HTTP.Paths {
			pathReg, err := regexp.Compile(path.Path)
			if err != nil {
				continue
			}
			backendObj := path.Backend
			s.host2Backends[rule.Host] = append(s.host2Backends[rule.Host], &backend{
				path: pathReg,
				url: &url.URL{
					Scheme: "http",
					Host:   fmt.Sprintf("%s:%s", backendObj.ServiceName, backendObj.ServicePort.String()),
				},
			})
		}
	}

}

func (s *Server) onIngressDelete(obj interface{}) {
	s.Lock()
	defer s.Unlock()
	ingress, ok := obj.(*v1beta1.Ingress)
	if !ok {
		return
	}
	spec := ingress.Spec
	if spec.Backend != nil {
		s.defaultURL = nil
	}
	for _, rule := range spec.Rules {
		if rule.HTTP == nil {
			continue
		}
		delete(s.host2Backends, rule.Host)
	}
}
