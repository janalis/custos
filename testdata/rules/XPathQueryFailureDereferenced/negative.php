<?php
function f(DOMXPath $x, $expr) { $nodes=$x->query($expr); if($nodes!==false){echo $nodes->length;} }
