<?php
function removeFile($p) { <warning descr="Return the deletion result or handle failure.">unlink($p)</warning>; return true; }
