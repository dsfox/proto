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

package echoclient

import (
	"context"

	api "github.com/teamgram/proto/v2/rpc/examples/echo/echo"
	"github.com/teamgram/proto/v2/rpc/examples/echo/echo/echo"
)

type EchoClient interface {
	EchosEcho(ctx context.Context, req *api.TLEchoEcho) (r *api.Echo, err error)
}

type defaultEchoClient struct {
	cli echo.Client
}

func NewEchoClient(cli echo.Client) EchoClient {
	return &defaultEchoClient{
		cli: cli,
	}
}

// EchosEcho
// echos.echo message:string = Echo;
func (m *defaultEchoClient) EchosEcho(ctx context.Context, in *api.TLEchoEcho) (*api.Echo, error) {
	return m.cli.EchoEcho(ctx, in)
}
