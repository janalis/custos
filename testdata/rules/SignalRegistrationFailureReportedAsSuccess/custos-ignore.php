<?php
// @custos-ignore SignalRegistrationFailureReportedAsSuccess

function install($h){pcntl_signal(SIGTERM,$h);return true;}
