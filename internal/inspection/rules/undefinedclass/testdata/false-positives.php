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
