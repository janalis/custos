<?php
// @noinspection SignalRegistrationFailureReportedAsSuccess

function install($h){pcntl_signal(SIGTERM,$h);return true;}
