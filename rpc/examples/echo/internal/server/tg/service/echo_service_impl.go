// Copyright (c) 2026 The Teamgram Authors (https://teamgram.net).
//  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"context"
	"encoding/json"

	"github.com/teamgram/proto/v2/rpc/examples/echo/echo"
	"github.com/teamgram/proto/v2/rpc/examples/echo/internal/core"

	"github.com/cloudwego/kitex/pkg/klog"
)

// EchoEcho
// echo.echo message:string = Echo;
func (s *Service) EchoEcho(ctx context.Context, request *echo.TLEchoEcho) (*echo.Echo, error) {
	c := core.New(ctx, s.svcCtx)
	klog.Infof("echos.echo - metadata: {}, request: %v", request)

	r, err := c.EchoEcho(request)
	if err != nil {
		return nil, err
	}

	txt, _ := json.Marshal(r)
	klog.Infof("echos.echo - reply: %s", string(txt))
	return r, err
}
