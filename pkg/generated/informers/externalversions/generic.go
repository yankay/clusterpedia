// Code generated by informer-gen. DO NOT EDIT.

package externalversions

import (
	"fmt"

	v1alpha2 "github.com/clusterpedia-io/api/cluster/v1alpha2"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	cache "k8s.io/client-go/tools/cache"
)

// GenericInformer is type of SharedIndexInformer which will locate and delegate to other
// sharedInformers based on type
type GenericInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() cache.GenericLister
}

type genericInformer struct {
	informer cache.SharedIndexInformer
	resource schema.GroupResource
}

// Informer returns the SharedIndexInformer.
func (f *genericInformer) Informer() cache.SharedIndexInformer {
	return f.informer
}

// Lister returns the GenericLister.
func (f *genericInformer) Lister() cache.GenericLister {
	return cache.NewGenericLister(f.Informer().GetIndexer(), f.resource)
}

// ForResource gives generic access to a shared informer of the matching type
// TODO extend this to unknown resources with a client pool
func (f *sharedInformerFactory) ForResource(resource schema.GroupVersionResource) (GenericInformer, error) {
	switch resource {
	// Group=cluster.clusterpedia.io, Version=v1alpha2
	case v1alpha2.SchemeGroupVersion.WithResource("clustersyncresources"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Cluster().V1alpha2().ClusterSyncResources().Informer()}, nil
	case v1alpha2.SchemeGroupVersion.WithResource("pediaclusters"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Cluster().V1alpha2().PediaClusters().Informer()}, nil

	}

	return nil, fmt.Errorf("no informer found for %v", resource)
}
