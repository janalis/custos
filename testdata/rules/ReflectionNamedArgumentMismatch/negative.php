<?php
function welcome($name) {} $r = new ReflectionFunction("welcome"); $r->invokeArgs(["name" => "Kai"]);
