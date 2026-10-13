<?php
function f($c) { $t=pg_escape_identifier($c,'inventory'); pg_query($c,"SELECT * FROM $t"); }
