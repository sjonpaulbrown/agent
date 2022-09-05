// Copyright 2022 Metrika Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package global

import (
	"context"
	"sync"
	"time"

	"agent/api/v1/model"

	"go.uber.org/zap"
)

// DefaultExporterRegisterer default exporter registerer to use for
// registering exporter implementations (see contrib package).
var DefaultExporterRegisterer = new(ExporterRegisterer)

// DefaultExporterTimeout exporter channel send timeout
// TODO: make timeout configurable per exporter basis
var DefaultExporterTimeout = 5 * time.Second

// Exporter interface describes the interface to be implemented for accessing
// the data stream generated by the enabled agent watchers.
type Exporter interface {
	// HandleMessage optionally processes and then exports an
	// incoming Metrika Agent Message (Metric or Event).
	// Used as a callback function by ExporterRegisterer on
	// every new message emitted by agent watchers.
	HandleMessage(ctx context.Context, msg *model.Message)
}

// ExporterHandler is the registerer's subscription unit.
type ExporterHandler struct {
	exporter       Exporter
	subscriptionCh <-chan interface{}
}

// ExporterRegisterer exporter handlers registry.
type ExporterRegisterer struct {
	handlers []ExporterHandler
}

// Register registers a new exporter and its channel.
func (e *ExporterRegisterer) Register(exporter Exporter, subCh chan interface{}) error {
	e.handlers = append(e.handlers, ExporterHandler{exporter: exporter, subscriptionCh: subCh})

	return nil
}

// Start starts a goroutine for each configured handler.
func (e *ExporterRegisterer) Start(ctx context.Context, wg *sync.WaitGroup) error {
	for i := range e.handlers {
		wg.Add(1)
		go func(e ExporterHandler) {
			MessageListener(ctx, wg, e.subscriptionCh, e.exporter)
		}(e.handlers[i])
	}

	return nil
}

// MessageListener reads from one Watcher emit channel
// and sequentially passes received messages to the exporter's
// HandleMessage method.
func MessageListener(ctx context.Context, wg *sync.WaitGroup, ch <-chan interface{}, e Exporter) {
	defer wg.Done()
	for {
		select {
		case m := <-ch:
			message, ok := m.(*model.Message)
			if !ok {
				zap.S().Warnf("Unexpected type %T, skipping item", m)
				continue
			}
			ctx, cancel := context.WithTimeout(ctx, DefaultExporterTimeout)
			e.HandleMessage(ctx, message)
			cancel()
		case <-ctx.Done():
			zap.S().Info("exiting listener")
			return
		}
	}
}
