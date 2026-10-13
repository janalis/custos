<?php
function f(XMLWriter $w) { <warning descr="Write untrusted XML content as escaped text.">$w->writeRaw($_POST['caption'])</warning>; }
