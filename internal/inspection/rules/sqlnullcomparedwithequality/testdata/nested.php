<?php
function nested(PDO $p){
    $p->query(<warning descr="Use IS NULL or IS NOT NULL for SQL null checks.">"SELECT id FROM t WHERE id IN (SELECT id FROM t) AND x = NULL"</warning>);
    $p->query(<warning descr="Use IS NULL or IS NOT NULL for SQL null checks.">"SELECT id FROM t WHERE EXTRACT(YEAR FROM d) = 2020 AND x = NULL"</warning>);
    $p->query(<warning descr="Use IS NULL or IS NOT NULL for SQL null checks.">"SELECT id FROM t WHERE a IS DISTINCT FROM b AND x = NULL"</warning>);
    $p->query(<warning descr="Use IS NULL or IS NOT NULL for SQL null checks.">"SELECT id FROM t WHERE (x = NULL)"</warning>);
    $p->query(<warning descr="Use IS NULL or IS NOT NULL for SQL null checks.">"SELECT id FROM t WHERE id IN (SELECT id FROM u WHERE y = NULL)"</warning>);
    $p->query(<warning descr="Use IS NULL or IS NOT NULL for SQL null checks.">"SELECT id FROM t WHERE x IS NOT NULL = NULL"</warning>);
    $p->query(<warning descr="Use IS NULL or IS NOT NULL for SQL null checks.">"SELECT id FROM t WHERE NOT NULL = x"</warning>);
    $p->query(<warning descr="Use IS NULL or IS NOT NULL for SQL null checks.">"SELECT id FROM t WHERE a = 1) AND x = NULL"</warning>);
}
