<?php
$pdo=new PDO("sqlite::memory:");$pdo->beginTransaction();<warning descr="Finish the transaction before beginning another one.">$pdo->beginTransaction()</warning>;
function assigned() {
    $p = new PDO("sqlite::memory:");
    $p->beginTransaction();
    $ok = (<warning descr="Finish the transaction before beginning another one.">$p->beginTransaction()</warning>);
    return $ok;
}
function tested() {
    $p = new PDO("sqlite::memory:");
    $p->beginTransaction();
    if (!<warning descr="Finish the transaction before beginning another one.">$p->beginTransaction()</warning>) {
        return false;
    }
    return true;
}
function returned() {
    $p = new PDO("sqlite::memory:");
    $p->beginTransaction();
    return <warning descr="Finish the transaction before beginning another one.">$p->beginTransaction()</warning> ? 1 : 0;
}
