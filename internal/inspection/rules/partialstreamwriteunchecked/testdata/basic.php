<?php
function bad($socket, $payload) { <warning descr="Verify that every payload byte was written.">fwrite($socket, $payload)</warning>; return true; }
