<?php
function bad($pattern) { foreach (<warning descr="Handle glob failure before iterating its result.">glob($pattern)</warning> as $path) { echo $path; } }
