<?php
function f(DOMXPath $x, $expr) { $nodes=$x->query($expr); if($nodes!==false){echo $nodes->length;} }
function truthy(DOMXPath $x, $expr) { $nodes=$x->query($expr); if(!$nodes){return 0;} $seen=0; return $nodes->length+$seen; }
function typed(DOMXPath $x, $expr) { $nodes=$x->query($expr); return $nodes instanceof DOMNodeList && $nodes->length>0; }
function inline(DOMXPath $x, $expr) { if(($nodes=$x->query($expr))===false){return 0;} return $nodes->length; }
function chosen(DOMXPath $x, $expr) { $nodes=$x->query($expr); return $nodes!==false ? $nodes->length : 0; }
