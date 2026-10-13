<?php
namespace Audit;const SQLITE3_ASSOC=2;function f(\SQLite3Result $r){echo json_encode(<warning descr="Fetch named SQLite columns before JSON serialization.">$r->fetchArray()</warning>);}
