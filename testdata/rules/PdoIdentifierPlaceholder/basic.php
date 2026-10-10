<?php
function bad(PDO $pdo) { <error descr="Select SQL identifiers from a trusted allowlist.">$pdo->prepare('SELECT * FROM :table')</error>; }
