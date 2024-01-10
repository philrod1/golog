..
.. Copyright (c) 2019 AT&T Intellectual Property.
..
.. Copyright (c) 2019 Nokia.
..
..
.. Licensed under the Creative Commons Attribution 4.0 International
..
.. Public License (the "License"); you may not use this file except
..
.. in compliance with the License. You may obtain a copy of the License at
..
..
..     https://creativecommons.org/licenses/by/4.0/
..
..
.. Unless required by applicable law or agreed to in writing, documentation
..
.. distributed under the License is distributed on an "AS IS" BASIS,
..
.. WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
..
.. See the License for the specific language governing permissions and
..
.. limitations under the License.
..
.. This source code is part of the near-RT RIC (RAN Intelligent Controller) platform project (RICP).
..

Developer Guide
===============

Clone the golog git repository
--------------------------------------
.. code:: bash

 git clone "https://gerrit.o-ran-sc.org/r/com/golog"

Initialization
--------------

A new logger instance is created with InitLogger function. Process identity is given as a parameter.

Mapped Diagnostics Context
--------------------------

The MDCs are key-value pairs, which are included to all log entries by the library.
The MDC pairs are logger instance specific.

The idea of the MDC is to define values, which stay the same over multiple log writings.
An MDC value set once will appear in all the subsequent logs written with the logger instance.

A logger instance can be shared by several goroutines.
Respectively, also the MDC values of the logger instance are then shared by them.
When sharing of the MDCs is not desired, separate logger instances should be used.

Log entry format
----------------

Each log entry written the library contains

 * Timestamp
 * Logger identity
 * Log entry severity
 * MDC pairs of the logger instance
 * Log message text

Currently the library only supports JSON formatted output written to standard out of the process

*Example log output*

`{"ts":1551183682974,"crit":"INFO","id":"myprog","mdc":{"second key":"other value","mykey":"keyval"},"msg":"hello world!"}`

 `{"ts":1602081593063,"crit":"INFO","id":"myapp","mdc":{"PID":"21587","POD_NAME":"tta-app-5565fc4d6f-ppfl8","CONTAINER_NAME":"tta-app","SERVICE_NAME":"TEST_APP","HOST_NAME":"master-an","SYSTEM_NAME":"CloudSpace-0"},"msg":"This is an example log"}`

Example
-------

.. code:: bash

 package main

 import (
        mdcloggo "gerrit.o-ran-sc.org/r/com/golog"
 )

 func main() {
    logFileMonitor := 0;
    logger, _ := mdcloggo.InitLogger("myname")
    if(logger.Mdclog_format_initialize(logFileMonitor)!=0) {
        logger.Error("UnSuccessful Format Initialize")
    }
    logger.MdcAdd("mykey", "keyval")
    logger.Info("Some test logs")
 } 

Logging Levels
--------------

.. code:: bash

 // ERR is an error level log entry.
   ERR Level = 1
 // WARN is a warning level log entry.
   WARN Level = 2
 // INFO is an info level log entry.
   INFO Level = 3
 // DEBUG is a debug level log entry.
   DEBUG Level = 4

Golog API's
-----------

1. LevelSet sets the current logging level.

.. code:: bash

 func (l *MdcLogger) LevelSet(level Level) 


2. LevelGet returns the current logging level.

.. code:: bash

 func (l *MdcLogger) LevelGet() Level

3. MdcAdd adds a new MDC key value pair to the logger.

.. code:: bash

 func (l *MdcLogger) MdcAdd(key string, value string)

4. MdcRemove removes an MDC key from the logger.

.. code:: bash

 func (l *MdcLogger) MdcRemove(key string)

5. MdcGet gets the value of an MDC from the logger.

.. code:: bash

 func (l *MdcLogger) MdcGet(key string) (string, bool)

Description: The function returns the value string and a boolean which tells if the key was found or not.

6. MdcClean removes all MDC keys from the logger.

.. code:: bash

 func (l *MdcLogger) MdcClean()

7. Mdclog_format_initialize Adds the MDC log format with HostName, PodName, ContainerName, ServiceName,PID,CallbackNotifyforLogFieldChange

.. code:: bash

 func (l *MdcLogger) Mdclog_format_initialize(log_change_monitor int) (int)

Description:  This api Initialzes mdclog print format using MDC Array by extracting the environment variables in the calling process for "SYSTEM_NAME", "HOST_NAME", "SERVICE_NAME", "CONTAINER_NAME", "POD_NAME" & "CONFIG_MAP_NAME"  mapped to HostName, ServiceName, ContainerName, Podname and Configuration-file-name of the services respectively.

  Note: In K8s/Docker Containers the environment variables are declared in the Helm charts.

  Refer xAPP developer guide for more information about how to define Helm chart.
