<?php
// @custos-ignore SelectWatchArraysNotRestored
$read=[$a,$b]; $write=null; $except=null; while ($running) { stream_select($read,$write,$except,1); }
