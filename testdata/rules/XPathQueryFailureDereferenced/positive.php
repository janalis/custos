<?php
function f(DOMXPath $x, $expr) { $nodes=$x->query($expr); echo <warning descr="Check the XPath query result before accessing its nodes.">$nodes->length</warning>; }
