<?php
function bad($path, $payload) { <warning descr="Verify the write before reporting success.">file_put_contents($path, $payload)</warning>; return true; }
