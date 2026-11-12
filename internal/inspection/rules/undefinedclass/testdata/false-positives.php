<?php
namespace App;

use App\Model\Member;
use Monolog\Logger;
use Sentry\EventException;

// Declared through a literal class_alias() in another file.
$member = new Member();

// Classes of PHP and of extensions the stubs describe, at any version.
$list = new \SplObjectStorage();
$frame = new \Imagick();
$old = new \HttpRequest();
// A global name whose leading word no builtin class uses: an extension the
// stubs lack.
$cache = new \Yac();
$search = new \SphinxClient();
// A namespace an extension uses may hold classes the stubs lack.
$doc = new \MongoDB\BSON\Document();

// Names PHP does not load.
function handle(\Throwable $e, Logger $logger): ?EventException
{
    if ($e instanceof EventException) {
        return null;
    }
    try {
        throw $e;
    } catch (\Vendor\Optional\Failure $failure) {
        return null;
    }
}
$class = \Vendor\Optional\Client::class;

#[\Vendor\Optional\Route('/home')]
class Controller {}

/** @param \Vendor\Optional\Request $request */
function docOnly($request) {}

// Guarded by name anywhere in the file.
if (class_exists(\Vendor\Optional\Cache::class)) {
    $cache = new \Vendor\Optional\Cache();
}
if (!interface_exists('Vendor\Optional\Contract')) {
    return;
}
class Adapter implements \Vendor\Optional\Contract {}

// Inside a condition about existence, extensions or versions.
if (extension_loaded('acme')) {
    $acme = new \AcmeHandle();
}
if (PHP_VERSION_ID >= 90000) {
    $future = new \FutureClass();
    $later = new \Php9\Later();
}

// Declared conditionally.
if (method_exists(\Vendor\Bridge\Kernel::class, 'boot')) {
    class Bridge extends \Vendor\Bridge\Kernel {}
}

// Dynamic names.
$name = 'Missing';
$object = new $name();
$name::create();

// self, static and parent.
class Base { public static function make() { return new static(); } }
class Child extends Base { public static function make() { return parent::make() ?? new self(); } }

// Guarded by an early exit.
function nativeHandle()
{
    if (!extension_loaded('acme')) {
        return null;
    }
    return new \Acme\Native\Handle();
}
function futureFeature()
{
    if (PHP_VERSION_ID < 90000) {
        throw new \LogicException('PHP 9 needed');
    }
    return \Php9\Feature::create();
}
function boot()
{
    defined('ACME_READY') or die();
    \Acme\Ready\Kernel::boot();
}
function bridge()
{
    if (!method_exists(\Vendor\Bridge\Loader::class, 'load')) {
        return;
    }
    \Vendor\Bridge\Loader::load();
}

// Inside a switch or match arm about extensions.
switch (true) {
    case extension_loaded('acme'):
        $switched = new \Acme\Switched\Handle();
        break;
}
$matched = match (true) {
    extension_loaded('acme') => new \Acme\Matched\Handle(),
    default => null,
};
