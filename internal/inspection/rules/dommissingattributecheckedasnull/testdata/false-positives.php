<?php
function f(DOMElement $e) { if (!$e->hasAttribute('key')) {} }
function g(?DOMElement $e) { return $e?->getAttribute('key') === null; }
function h(?DOMElement $e) { $v = $e?->getAttribute('key'); return $v !== null; }
