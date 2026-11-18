<?php
function f(DOMXPath $x, $expr) { $nodes=$x->query($expr); echo <warning descr="Check the XPath query result before accessing its nodes.">$nodes->length</warning>; }
function inverted(DOMXPath $x, $expr) { $nodes=$x->query($expr); if(!$nodes){return <warning descr="Check the XPath query result before accessing its nodes.">$nodes->length</warning>;} return 0; }
function direct(DOMXPath $x, $expr) { return <warning descr="Check the XPath query result before accessing its nodes.">$x->query($expr)->item(0)</warning>; }
