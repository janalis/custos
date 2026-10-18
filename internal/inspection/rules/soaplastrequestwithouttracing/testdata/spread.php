<?php

function optionsBySpread(array $options): ?string
{
    $client = new SoapClient(null, ...$options);

    return $client->__getLastRequest();
}
