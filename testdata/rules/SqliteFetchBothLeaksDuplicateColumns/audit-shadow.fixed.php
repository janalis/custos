<?php
namespace Audit;const SQLITE3_ASSOC=2;function f(\SQLite3Result $r){echo json_encode($r->fetchArray(\SQLITE3_ASSOC));}
