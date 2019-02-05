package updater

import (
	"github.com/filintod/mapupdatego/prefix"
	"reflect"
)

/*
This package tries to create topological sorting of files that will end up as maps (json/yaml) like map([string]interface)

Special Prefixes:
- (+) APPEND_PREFIX: when found on a Map or slice means to add to parent Map if not found on Parent.  On Map this is actually
					 the default behavior, but on Slices the default behavior is to replace the parent slice with the current
					 node slice. On Slices it will only append if value is not found.
- (++) APPEND_ALL_PREFIX: This is similar to APPEND_PREFIX, but on slices it will also allow to append even if the value is repeated
- (-) REMOVE_PREFIX: when found on a Map or Slice means to remove the key on Maps or the first value on Slice and not carry to children nodes.
					 on slices
- (--) REMOVE_ALL_PREFIX: when found on a Map removes the key and on a slice removes all occurrences of the value
- (index) INDEX_PREFIX: on a Map this is not used but on a slice it means that we want to put it at a certain location.
					  the index ({index}) is an integer. A negative value will be used as an index from the end (-1 is the last element)
 					  The index is zero based, so the value 0 is the first element.
# index is not implemented yet
- (index+) INDEX_APPEND_PREFIX: append to parent and put at certain location for slices only. will not append if the value already present
- (index++) INDEX_APPEND_ALL_PREFIX: append to parent and put at certain location for slices only and repeated values are ok


Map will be coalesce from their parent.  There is an implicit APPEND_PREFIX on maps, so if a key is found on a child
node but not on the parent it will be added. If a key is found on both the child value will take precedence unless
the value is also a map or slice where it will be recursively coalesce.

A special REMOVE_PREFIX, by default (-), will make the key to be deleted.
For example if parent map has a key "a" and current node has a key "(-)a" then this key will not be cascaded down
current node's children or if current node is a leaf this key will not be shown on the final map.

For slices, there are three possible cases:

- Slices of strings ([]string). If there is one value in the slice with one of the special prefixes (APPEND_PREFIX, by default (+), or REMOVE_PREFIX)
	then all elements of the current node slice will be inheriting from its parent and by default all values without a prefix
	will be set an APPEND_PREFIX.

- For slice of maps, the APPEND or REMOVE PREFIX has to be put on the key(s) you want to use to compare with. For example:
	Parent node:
	- name: peter
	  last: pan
      value: v1
	- name: mickey
	  last: mouse
      value: v2

	Current node:
	- (-)name: peter

	In this case the final slice will have the first element (name: peter, last: pan) removed once
	The values should be comparable (https://golang.org/ref/spec#Comparison_operators). Multi-keys area allowed and they
	will be ANDed to compare.  Mismatch on (-) and (+) is not allowed and a critical error will be issued

- Slices of scalar (int, float, bool, etc). In this case they are just replaced by current node if present.
*/

// TODO: implement indexed prefixes
// TODO: validate value with unknown prefix
// compAndSetSliceStr check each element of childVal and parentVal slices to see if elements in parent are present on
//	the child and viceversa. if an element in child is prefix with - we don't copy the value from parent if present
//	the algo creates a new list with all elements in childVal that are not present in parentVal plus all element in
// 	parentVal not present in childVal
func coalesceStrSlice(childVal, parentVal reflect.Value) {
	newSlice := make([]string, 0)
	parentValSlice := make([]string, 0)
	parentValKeys := make(map[string][]int, parentVal.Len())

	// creates set of parentVal Keys not including those prefixed with REMOVE/REMOVEALL PREFIXes value
	for i := 0; i < parentVal.Len(); i++ {
		pv := parentVal.Index(i).String()
		if !(PREFIX.HasRemove(pv) || PREFIX.HasRemoveAll(pv)) {
			parentValSlice = append(parentValSlice, pv)
			if _, ok := parentValKeys[pv]; ok {
				parentValKeys[pv] = append(parentValKeys[pv], len(parentValSlice))
			} else {
				parentValKeys[pv] = []int{len(parentValSlice)}
			}
		}
	}

	// if any element on current (child) node has a special prefix then we are inheriting
	isInheriting := false
	setInherit := func(cv string) string {
		if !isInheriting { // if first time and we have something in newSlice (those without prefix) copy those to parent
			for _, v := range newSlice {
				if _, ok := parentValKeys[v]; !ok {
					parentValSlice = append(parentValSlice, v)
					parentValKeys[v] = []int{len(parentValSlice)}
				}
			}
			newSlice = nil
		}
		isInheriting = true
		return cv
	}

	for i := 0; i < childVal.Len(); i++ {
		cv := childVal.Index(i).String()
		switch {
		case PREFIX.HasRemove(cv):
			cv = setInherit(PREFIX.TrimRemove(cv))
			if _, ok := parentValKeys[cv]; ok {
				parentValKeys[cv] = parentValKeys[cv][1:]
				if len(parentValKeys[cv]) == 0 {
					delete(parentValKeys, cv)
				}
			}
		case PREFIX.HasRemoveAll(cv):
			cv = setInherit(PREFIX.TrimRemoveAll(cv))
			if _, ok := parentValKeys[cv]; ok {
				delete(parentValKeys, cv)
			}
		case PREFIX.HasAppend(cv):
			cv = setInherit(PREFIX.TrimAppend(cv))
			if _, ok := parentValKeys[cv]; !ok {
				parentValSlice = append(parentValSlice, cv)
				parentValKeys[cv] = []int{len(parentValSlice)}
			}

		case PREFIX.HasAppendAll(cv):
			cv = setInherit(PREFIX.TrimAppendAll(cv))
			if _, ok := parentValKeys[cv]; ok {
				parentValKeys[cv] = append(parentValKeys[cv], len(parentValSlice))
			} else {
				parentValKeys[cv] = []int{len(parentValSlice)}
			}

		case PREFIX.HasIndex(cv):
			// TODO: implement

		case PREFIX.HasIndexAppend(cv):
			// TODO: implement

		case PREFIX.HasIndexAppendAll(cv):
			// TODO: implement

		case isInheriting: // here we don't have a prefix but we have had in the past
			if _, ok := parentValKeys[cv]; !ok {
				parentValSlice = append(parentValSlice, cv)
				parentValKeys[cv] = []int{len(parentValSlice)}
			}

		default:
			newSlice = append(newSlice, cv)
		}
	}

	var slice []string
	if isInheriting {
		slice = parentValSlice
	} else {
		slice = newSlice
	}
	childVal.Set(reflect.MakeSlice(parentVal.Type(), len(slice), len(slice)))
	reflect.Copy(childVal, reflect.ValueOf(slice))
}

// coalesceMap is similar to the compAndSetSliceStr where we add elements not found in child that are found in parent
//	  Here we also allowed keys to be prefixed with REMOVEPREFIX (dash) and when that happen the parent key is not used by the children
func coalesceMap(current, parent reflect.Value, skipKeys map[string]bool) {

	parentKeys := parent.MapKeys()
	parentValKeys := make(map[reflect.Value]reflect.Value, len(parentKeys))

	for i := 0; i < len(parentKeys); i++ {
		if !PREFIX.HasRemove(parentKeys[i].String()) {
			parentValKeys[parentKeys[i]] = parent.MapIndex(parentKeys[i])
		}
	}

	if current.IsNil() && len(parentValKeys) != 0 {
		current.Set(reflect.MakeMap(parent.Type()))
	} else {
		for _, k := range current.MapKeys() {
			if PREFIX.HasRemove(k.String()) {
				delete(parentValKeys, reflect.ValueOf(PREFIX.TrimRemove(k.String())))

			} else if pv, ok := parentValKeys[k]; ok {
				// do a recursive parse (coalesce) on the map value
				coalesce(current.MapIndex(k), pv, skipKeys)
				delete(parentValKeys, parent.MapIndex(k))
			}
		}
	}

	// the remaining element on parent not found on current are coalesced down
	for k, v := range parentValKeys {
		current.SetMapIndex(k, v)
	}
}

// compAndSetSlices replaces parent slice with child slice if present.  If the slice is of type str
// 	  then we do a compare and append if value not present by using func compAndSetSliceStr
func coalesceSlices(childVal, parentVal reflect.Value) {
	switch childVal.Type().Elem().Kind() {
	case reflect.String:
		// we do append if missing for string slices and also remove prefixed - values
		coalesceStrSlice(childVal, parentVal)
	case reflect.Map:
		// TODO: implement slice of maps
	default:
		if childVal.Len() == 0 && parentVal.Len() != 0 {
			// we need to expand/contract child to the len/cap of parent as we are doing a complete replacement
			childVal.Set(reflect.MakeSlice(parentVal.Type(), parentVal.Len(), parentVal.Cap()))
			reflect.Copy(childVal, parentVal)
		}
	}
}

// coalesce recursively goes through every element not present in child but present in parent and copies it
func coalesce(childVal, parentVal interface{}, skipKeys map[string]bool) {

	switch childVal.(type) {

	case map[string]interface{}:
		print("MAP: %v", childVal)

	case []string:
		print("STR list: %v", childVal)

	case []int, []float64, []bool:
		print("list: %v", childVal)

	case string, int, bool, float64:
		print("Scalar: %v", childVal)
	default:
		v := reflect.ValueOf(childVal)
		k := v.Kind()
		print(k)
	}
}

// RenderCharts merges the properties of parents onto their children (in place)
//func SortDependencies(maps []map[string]interface{}, omitFields map[string]bool, profileType string) {
//	//
//	parentChildren := make(map[string][]string)
//	roots := make([]map[string]interface{}, 0)
//	visited := make(map[string]bool)
//
//	// create the parent to children map
//	for _, info := range maps {
//		// roots are the ones without parent (@bases not found or @bases=[])
//		parents := info["@bases"]
//		if parents == nil || len(parents.([]string)) == 0 {
//			roots = append(roots, info)
//			continue
//		}
//		for _, parent := range parents.([]string) {
//			if _, ok := parentChildren[parent]; !ok {
//				parentChildren[parent] = make([]string, 0)
//			}
//			parentChildren[parent] = append(parentChildren[parent], info.GetName())
//		}
//
//	}
//
//	// we go from roots down as this is a simpler structure where there is only one parent per child
//	// and append children to roots slice and to visited to avoid duplicated work
//	var parentProfile string
//
//	for len(roots) != 0 {
//		// remove top parent from roots and shift roots to the left
//		parentProfile, roots = roots[0], roots[1:]
//		p := reflect.ValueOf(s.getParentProfile(parentProfile, profileType)).Elem()
//		for _, childProfile := range parentChildren[parentProfile] {
//			c := reflect.ValueOf(s.getParentProfile(childProfile, profileType)).Elem()
//			coalesce(c, p, omitFields)
//
//			if !visited[childProfile] {
//				roots = append(roots, childProfile)
//				visited[childProfile] = true
//			}
//		}
//	}
//	// now that everything is correct set the wasRendered flag to true
//	for hp := range s.iterateProfiles(profileType) {
//		hp.SetRender()
//	}
//}

var PREFIX = prefix.NewPrefix()

var REGEX_HELPER = `{{\/\*[\s\S]+?\*\/}}\s+{{-? define "(?P<tname>[^"]+)" [\s\S]+?{{\-?\s+end\s+-?}}\n{2}`
