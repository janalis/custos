<?php
function f($c) { $v=pg_escape_literal($c,'Bex'); pg_query($c,"SELECT $v"); }
