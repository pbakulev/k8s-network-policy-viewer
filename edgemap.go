package main

import (
)

func initializeEdgeMap(edgeMap *map[string][]string, namespacePodMap *map[string][]string) {
	var allPods []string
	for _, v := range *namespacePodMap {
		for _, s := range v {
			allPods = append(allPods, s)
		}
	}
	for _, outer := range allPods {
		for _, inner := range allPods {
			if inner == outer {
				continue
			}
			(*edgeMap)[outer] = append((*edgeMap)[outer], inner)
		}
	}
}

func filterEdgeMap(edgeMap *map[string][]string, namespacePodMap *map[string][]string, podLabelMap *map[string]map[string]string, networkPolicies *[]ApiObject) {
	for _, o := range *networkPolicies {
		podsSet := make(map[string]struct{})
		namespace := o.Metadata.Namespace
		for _, pod := range (*namespacePodMap)[namespace] {
			podsSet[pod] = struct{}{}
		}
		// 1. apply blanket ingress/egress policies
		if len(o.Spec.PolicyTypes) == 0 {
			// if none specified, default to ingress policy
			filterIngress(&podsSet, edgeMap)
		} else {
			for _, policyType := range o.Spec.PolicyTypes {
				switch policyType {
				case "Ingress":
					filterIngress(&podsSet, edgeMap)
				case "Egress":
					filterEgress(&podsSet, edgeMap)
				}
			}
		}
		// 2. now deal with whitelisted pods
		//selectedPods := selectPods(namespace, &o.Spec.PodSelector.MatchLabels, namespacePodMap, podLabelMap)
	}
}

// TODO: apply filter only once when multiple network policies apply to one namespace
func filterIngress(podsSet *map[string]struct{}, edgeMap *map[string][]string) {
	for fromString, toSlice := range *edgeMap {
		for i, pod := range toSlice {
			if _, ok := (*podsSet)[pod]; ok {
				// pod in scope, so remove from edgeMap
				toSlice = append(toSlice[:i], toSlice[i+1:]...)
				(*edgeMap)[fromString] = toSlice
			}
		}
	}
}

// support among SDN providers still patchy?
// examine desired not necessarily actual state
// TODO: as with ingress, apply filter only once
func filterEgress(podsSet *map[string]struct{}, edgeMap *map[string][]string) {
	for pod, _ := range *podsSet {
		(*edgeMap)[pod] = nil
	}
}

func unique(slice []string) []string {
	keys := make(map[string]struct{})
	list := []string{}
	for _, entry := range slice {
		if _, ok := keys[entry]; !ok {
			keys[entry] = struct{}{}
			list = append(list, entry)
		}
	}
	return list
}
