<?php

function describe($entity = null, $label = 'x', ?Order $order = null, Invoice $invoice = NULL) {
    <warning descr="get_class() rejects null on PHP 7.2+; guard the argument.">\get_class(NULL)</warning>;
    <warning descr="get_class() rejects null on PHP 7.2+; guard the argument.">get_class($entity)</warning>;
    <warning descr="get_class() rejects null on PHP 7.2+; guard the argument.">get_class($order)</warning>;

    get_class($label);
    get_class($entity, 1);

    if (empty($invoice)) {
        return;
    }
    get_class($invoice);

    while ($order instanceof Order) {
        get_class($order);
    }
}

function report(?Order $first, ?Order $second, ?Order $third) {
    $tag = $first ? 'a' : 'b';
    echo get_class($first);
    echo <warning descr="get_class() rejects null on PHP 7.2+; guard the argument.">get_class($second)</warning>;
    if (null != $second) {}
    if (!$third) { return; }
    echo get_class($third);
}

$loose = null;
<warning descr="get_class() rejects null on PHP 7.2+; guard the argument.">get_class($loose)</warning>;
