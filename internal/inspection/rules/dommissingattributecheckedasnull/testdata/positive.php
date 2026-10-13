<?php
function f(DOMElement $e) { if (<warning descr="Use hasAttribute to test whether an attribute exists.">$e->getAttribute('key') === null</warning>) {} }
