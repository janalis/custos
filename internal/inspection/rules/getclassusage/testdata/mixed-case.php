<?php

function label(?object $item): string
{
    return <warning descr="get_class() rejects null on PHP 7.2+; guard the argument.">Get_Class($item)</warning>;
}
