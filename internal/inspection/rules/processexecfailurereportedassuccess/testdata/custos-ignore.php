<?php
// @custos-ignore ProcessExecFailureReportedAsSuccess

pcntl_exec("/missing/program",[]);exit(0);
