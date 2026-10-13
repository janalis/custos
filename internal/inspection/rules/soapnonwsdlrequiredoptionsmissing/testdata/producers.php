<?php
// A WSDL assigned on several paths is not known to be absent.

function wsdlOnOnePath($flag, $wsdl)
{
    if ($flag) {
        $source = null;
    } else {
        $source = $wsdl;
    }

    return new SoapClient($source, []);
}

function wsdlOnTheFirstPath($flag, $wsdl)
{
    if ($flag) {
        $source = $wsdl;
    } else {
        $source = null;
    }

    return new SoapClient($source, []);
}

function wsdlSetLater($flag)
{
    $source = null;
    if ($flag) {
        $source = 'https://example.net/service.wsdl';
    }

    return new SoapClient($source, []);
}

function wsdlSetInAnOperand($flag, $wsdl)
{
    $source = null;
    $flag && $source = $wsdl;

    return new SoapClient($source, []);
}

function wsdlFromALoop($list)
{
    $source = null;
    foreach ($list as $source) {
    }

    return new SoapClient($source, []);
}

function wsdlFromATry($finder)
{
    $source = null;
    try {
        $source = $finder->find();
    } catch (\Exception $e) {
    }

    return new SoapClient($source, []);
}

function noWsdlOnALaterPass($count, $wsdl)
{
    $source = $wsdl;
    while ($count--) {
        new SoapClient($source, []);
        $source = null;
    }
}

// The first pass of the loop has no WSDL.
function wsdlOnALaterPass($count, $wsdl)
{
    $source = null;
    while ($count--) {
        <warning descr="Supply location and uri in non-WSDL mode.">new SoapClient($source, [])</warning>;
        $source = $wsdl;
    }
}
