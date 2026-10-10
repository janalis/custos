<?php
function f($c) { $t=<warning descr="Escape PostgreSQL identifiers with the identifier API.">pg_escape_literal($c,'inventory')</warning>; pg_query($c,"SELECT * FROM $t"); }
