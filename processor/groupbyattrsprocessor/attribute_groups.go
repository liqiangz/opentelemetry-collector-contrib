// Copyright 2020 OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package groupbyattrsprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/groupbyattrsprocessor"

import (
	"go.opentelemetry.io/collector/model/pdata"
)

func instrumentationLibrariesEqual(il1, il2 pdata.InstrumentationLibrary) bool {
	return il1.Name() == il2.Name() && il1.Version() == il2.Version()
}

// matchingInstrumentationLibrarySpans searches for a pdata.InstrumentationLibrarySpans instance matching
// given InstrumentationLibrary. If nothing is found, it creates a new one
func matchingInstrumentationLibrarySpans(rl pdata.ResourceSpans, library pdata.InstrumentationLibrary) pdata.InstrumentationLibrarySpans {
	ilss := rl.InstrumentationLibrarySpans()
	for i := 0; i < ilss.Len(); i++ {
		ils := ilss.At(i)
		if instrumentationLibrariesEqual(ils.InstrumentationLibrary(), library) {
			return ils
		}
	}

	ils := ilss.AppendEmpty()
	library.CopyTo(ils.InstrumentationLibrary())
	return ils
}

// matchingInstrumentationLibraryLogs searches for a pdata.InstrumentationLibraryLogs instance matching
// given InstrumentationLibrary. If nothing is found, it creates a new one
func matchingInstrumentationLibraryLogs(rl pdata.ResourceLogs, library pdata.InstrumentationLibrary) pdata.InstrumentationLibraryLogs {
	ills := rl.InstrumentationLibraryLogs()
	for i := 0; i < ills.Len(); i++ {
		ill := ills.At(i)
		if instrumentationLibrariesEqual(ill.InstrumentationLibrary(), library) {
			return ill
		}
	}

	ill := ills.AppendEmpty()
	library.CopyTo(ill.InstrumentationLibrary())
	return ill
}

// matchingInstrumentationLibraryMetrics searches for a pdata.InstrumentationLibraryMetrics instance matching
// given InstrumentationLibrary. If nothing is found, it creates a new one
func matchingInstrumentationLibraryMetrics(rm pdata.ResourceMetrics, library pdata.InstrumentationLibrary) pdata.InstrumentationLibraryMetrics {
	ilms := rm.InstrumentationLibraryMetrics()
	for i := 0; i < ilms.Len(); i++ {
		ilm := ilms.At(i)
		if instrumentationLibrariesEqual(ilm.InstrumentationLibrary(), library) {
			return ilm
		}
	}

	ilm := ilms.AppendEmpty()
	library.CopyTo(ilm.InstrumentationLibrary())
	return ilm
}

// spansGroupedByAttrs keeps all found grouping attributes for spans, together with the matching records
type spansGroupedByAttrs struct {
	pdata.ResourceSpansSlice
}

// logsGroupedByAttrs keeps all found grouping attributes for logs, together with the matching records
type logsGroupedByAttrs struct {
	pdata.ResourceLogsSlice
}

// metricsGroupedByAttrs keeps all found grouping attributes for metrics, together with the matching records
type metricsGroupedByAttrs struct {
	pdata.ResourceMetricsSlice
}

func newLogsGroupedByAttrs() *logsGroupedByAttrs {
	return &logsGroupedByAttrs{
		ResourceLogsSlice: pdata.NewResourceLogsSlice(),
	}
}

func newSpansGroupedByAttrs() *spansGroupedByAttrs {
	return &spansGroupedByAttrs{
		ResourceSpansSlice: pdata.NewResourceSpansSlice(),
	}
}

func newMetricsGroupedByAttrs() *metricsGroupedByAttrs {
	return &metricsGroupedByAttrs{
		ResourceMetricsSlice: pdata.NewResourceMetricsSlice(),
	}
}

// Build the Attributes that we'll be looking for in existing Resources as a merge of the Attributes
// of the original Resource with the requested Attributes
func buildReferenceAttributes(originResource pdata.Resource, requiredAttributes pdata.AttributeMap) pdata.AttributeMap {
	referenceAttributes := pdata.NewAttributeMap()
	originResource.Attributes().CopyTo(referenceAttributes)
	requiredAttributes.Range(func(k string, v pdata.AttributeValue) bool {
		referenceAttributes.Upsert(k, v)
		return true
	})
	return referenceAttributes
}

// resourceMatches verifies if given pdata.Resource attributes strictly match with the specified
// reference Attributes (all attributes must match strictly)
func resourceMatches(resource pdata.Resource, referenceAttributes pdata.AttributeMap) bool {

	// If not the same number of attributes, it doesn't match
	if referenceAttributes.Len() != resource.Attributes().Len() {
		return false
	}

	// Go through each attribute and check the corresponding attribute value in the tested Resource
	matching := true
	referenceAttributes.Range(func(referenceKey string, referenceValue pdata.AttributeValue) bool {
		testedValue, foundKey := resource.Attributes().Get(referenceKey)
		if !foundKey || !referenceValue.Equal(testedValue) {
			// One difference is enough to consider it doesn't match, so fail early
			matching = false
			return false
		}
		return true
	})

	return matching
}

// findResource searches for an existing pdata.ResourceLogs that strictly matches with the specified reference
// Attributes. Returns the matching pdata.ResourceLogs and bool value which is set to true if found
func (lgba logsGroupedByAttrs) findResource(referenceAttributes pdata.AttributeMap) (pdata.ResourceLogs, bool) {
	for i := 0; i < lgba.Len(); i++ {
		if resourceMatches(lgba.At(i).Resource(), referenceAttributes) {
			return lgba.At(i), true
		}
	}
	return pdata.ResourceLogs{}, false
}

// findResource searches for an existing pdata.ResourceLogs that strictly matches with the specified reference
// Attributes. Returns the matching pdata.ResourceLogs and bool value which is set to true if found
func (sgba spansGroupedByAttrs) findResource(referenceAttributes pdata.AttributeMap) (pdata.ResourceSpans, bool) {
	for i := 0; i < sgba.Len(); i++ {
		if resourceMatches(sgba.At(i).Resource(), referenceAttributes) {
			return sgba.At(i), true
		}
	}
	return pdata.ResourceSpans{}, false
}

// findResource searches for an existing pdata.ResourceMetrics that strictly matches with the specified reference
// Attributes. Returns the matching pdata.ResourceMetrics and bool value which is set to true if found
func (mgba metricsGroupedByAttrs) findResource(referenceAttributes pdata.AttributeMap) (pdata.ResourceMetrics, bool) {

	for i := 0; i < mgba.Len(); i++ {
		if resourceMatches(mgba.At(i).Resource(), referenceAttributes) {
			return mgba.At(i), true
		}
	}
	return pdata.ResourceMetrics{}, false
}

// Update the specified (and new) Resource with the properties of the original Resource, and with the
// required Attributes
func updateResourceToMatch(newResource pdata.Resource, originResource pdata.Resource, requiredAttributes pdata.AttributeMap) {

	originResource.CopyTo(newResource)

	// This prioritizes required attributes over the original resource attributes, if they overlap
	attrs := newResource.Attributes()
	requiredAttributes.Range(func(k string, v pdata.AttributeValue) bool {
		attrs.Upsert(k, v)
		return true
	})

}

// findOrCreateResource searches for a Resource with matching attributes and returns it. If nothing is found, it is being created
func (sgba *spansGroupedByAttrs) findOrCreateResource(originResource pdata.Resource, requiredAttributes pdata.AttributeMap) pdata.ResourceSpans {

	// Build the reference attributes that we're looking for in Resources
	referenceAttributes := buildReferenceAttributes(originResource, requiredAttributes)

	// Do we have a matching Resource?
	resource, found := sgba.findResource(referenceAttributes)
	if found {
		return resource
	}

	// Not found: create a new resource
	resource = sgba.AppendEmpty()
	updateResourceToMatch(resource.Resource(), originResource, requiredAttributes)
	return resource

}

// findResourceOrElseCreate searches for a Resource with matching attributes and returns it. If nothing is found, it is being created
func (lgba *logsGroupedByAttrs) findResourceOrElseCreate(originResource pdata.Resource, requiredAttributes pdata.AttributeMap) pdata.ResourceLogs {

	// Build the reference attributes that we're looking for in Resources
	referenceAttributes := buildReferenceAttributes(originResource, requiredAttributes)

	// Do we have a matching Resource?
	resource, found := lgba.findResource(referenceAttributes)
	if found {
		return resource
	}

	// Not found: create a new resource
	resource = lgba.AppendEmpty()
	updateResourceToMatch(resource.Resource(), originResource, requiredAttributes)
	return resource

}

// findResourceOrElseCreate searches for a Resource with matching attributes and returns it. If nothing is found, it is being created
func (mgba *metricsGroupedByAttrs) findResourceOrElseCreate(originResource pdata.Resource, requiredAttributes pdata.AttributeMap) pdata.ResourceMetrics {

	// Build the reference attributes that we're looking for in Resources
	referenceAttributes := buildReferenceAttributes(originResource, requiredAttributes)

	// Do we have a matching Resource?
	resource, found := mgba.findResource(referenceAttributes)
	if found {
		return resource
	}

	// Not found: create a new resource
	resource = mgba.AppendEmpty()
	updateResourceToMatch(resource.Resource(), originResource, requiredAttributes)
	return resource

}
