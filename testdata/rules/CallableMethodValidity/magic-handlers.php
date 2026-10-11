<?php
class Relay
{
    private function fetch() {}
    private static function cache() {}
    public function load() {}
    public function __call($name, $args) {}
}

class Hub
{
    private function fetch() {}
    public function load() {}
    public static function __callStatic($name, $args) {}
}

class Mesh extends Relay
{
    public static function __callStatic($name, $args) {}
}

class Node extends Relay
{
    public function viaThis()
    {
        return [
            is_callable('Relay::fetch'),
            is_callable([parent::class, 'cache']),
        ];
    }

    public static function noThis()
    {
        return is_callable(<warning descr="Method 'fetch' is not public, so the callback cannot be invoked from outside."><warning descr="Method 'fetch' is not static but is referenced without an object.">'Relay::fetch'</warning></warning>);
    }
}

class Spoke extends Hub
{
    public function viaThis() { return is_callable(<warning descr="Method 'fetch' is not public, so the callback cannot be invoked from outside.">[Hub::class, 'fetch']</warning>); }
    public static function noThis() { return is_callable([Hub::class, 'fetch']); }
}

function handlers(Relay $relay, Hub $hub, Mesh $mesh)
{
    return [
        is_callable([$relay, 'fetch']),
        is_callable([$relay, 'cache']),
        is_callable(<warning descr="Method 'fetch' is not public, so the callback cannot be invoked from outside."><warning descr="Method 'fetch' is not static but is referenced without an object.">'Relay::fetch'</warning></warning>),
        is_callable(<warning descr="Method 'fetch' is not public, so the callback cannot be invoked from outside."><warning descr="Method 'fetch' is not static but is referenced without an object.">['Relay', 'fetch']</warning></warning>),
        is_callable('Hub::fetch'),
        is_callable([Hub::class, 'fetch']),
        is_callable(<warning descr="Method 'fetch' is not public, so the callback cannot be invoked from outside.">[$hub, 'fetch']</warning>),
        is_callable(<warning descr="Method 'load' is not static but is referenced without an object.">'Hub::load'</warning>),
        is_callable(<warning descr="Method 'load' is not static but is referenced without an object.">'Mesh::load'</warning>),
        is_callable('Mesh::fetch'),
        is_callable([$mesh, 'fetch']),
    ];
}
