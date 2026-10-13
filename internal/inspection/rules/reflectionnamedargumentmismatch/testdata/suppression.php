<?php
// @custos-ignore ReflectionNamedArgumentMismatch
function welcome($name) {} $r = new ReflectionFunction("welcome"); $r->invokeArgs(["nickname" => "Kai"]);
