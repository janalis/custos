<?php
function positionalTables(PDO $pdo) {
    $a = <error descr="Select SQL identifiers from a trusted allowlist.">$pdo->prepare('SELECT * FROM ?')</error>;
    $b = <error descr="Select SQL identifiers from a trusted allowlist.">$pdo->prepare('SELECT * FROM t JOIN ? ON 1 = 1')</error>;
    $c = <error descr="Select SQL identifiers from a trusted allowlist.">$pdo->prepare('UPDATE ? SET a = 1')</error>;
    $d = <error descr="Select SQL identifiers from a trusted allowlist.">$pdo->prepare('INSERT INTO ? VALUES (1)')</error>;
    $e = <error descr="Select SQL identifiers from a trusted allowlist.">$pdo->prepare('SELECT * FROM (SELECT * FROM :source) AS s')</error>;
    $f = <error descr="Select SQL identifiers from a trusted allowlist.">$pdo->prepare('SELECT EXTRACT(YEAR FROM created) FROM :source')</error>;
    return [$a, $b, $c, $d, $e, $f];
}
