<?php
function welcome($name) {} $r = new ReflectionFunction("welcome"); <error descr="Use a declared reflection parameter name.">$r->invokeArgs(["nickname" => "Kai"])</error>;
