<?php
function viaDefault()
{
    stream_context_set_default(['http' => ['ignore_errors' => true]]);
    return file_get_contents('http://service.test/missing');
}
function viaParams()
{
    $ctx = stream_context_create(['http' => ['method' => 'GET']], ['options' => ['http' => ['ignore_errors' => true]]]);
    return file_get_contents('http://service.test/missing', false, $ctx);
}
function viaUnknownParams(array $params)
{
    return fopen('http://service.test/missing', 'r', false, stream_context_create(['http' => ['method' => 'GET']], $params));
}
