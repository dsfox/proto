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

package server

import (
	"flag"
	"log"

	"github.com/teamgram/proto/v2/rpc/codec"
	"github.com/teamgram/proto/v2/rpc/examples/echo/echo/echo"
	"github.com/teamgram/proto/v2/rpc/examples/echo/internal/config"
	"github.com/teamgram/proto/v2/rpc/examples/echo/internal/server/tg/service"
	"github.com/teamgram/proto/v2/rpc/examples/echo/internal/svc"

	"github.com/cloudwego/kitex/server"
)

var configFile = flag.String("f", "etc/echo.yaml", "the config file")

type Server struct {
	server.Server
}

func New() *Server {
	return new(Server)
}

func (s *Server) Initialize() error {
	var c config.Config
	ctx := svc.NewServiceContext(c)
	_ = ctx

	cCodec := codec.NewZRpcCodec(true)
	s.Server = echo.NewServer(service.New(ctx), server.WithCodec(cCodec))
	return nil
}

func (s *Server) RunLoop() {
	if err := s.Server.Run(); err != nil {
		log.Println("server stopped with error:", err)
	} else {
		log.Println("server stopped")
	}
}

func (s *Server) Destroy() {
}
