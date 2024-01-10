/*
 *  Copyright (c) 2019 AT&T Intellectual Property.
 *  Copyright (c) 2018-2019 Nokia.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 *  This source code is part of the near-RT RIC (RAN Intelligent Controller)
 *  platform project (RICP).
 */

package main

import (
	"fmt"
	"os"
	"time"
	mdcloggo "gerrit.o-ran-sc.org/r/com/golog"
)

func main() {
	logger, _ := mdcloggo.InitLogger("myname")
	logFileMonitor := 0
	logger.MdcAdd("foo", "bar")
	logger.MdcAdd("foo2", "bar2")
	if logger.Mdclog_format_initialize(logFileMonitor) != 0 {
		logger.Error("Failed in MDC Log Format Initialize")
	}

	start := time.Now()
	for i := 0; i < 10; i++ {
		logger.Info("Some test logs")
	}
	elapsed := time.Since(start)
	fmt.Fprintf(os.Stderr, "Elapsed %v\n", elapsed)
}
