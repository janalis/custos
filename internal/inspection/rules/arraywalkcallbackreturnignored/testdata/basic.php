<?php
array_walk($names,<warning descr="Use array_map to retain the callback results.">fn(string $name)=>strtoupper($name)</warning>);
