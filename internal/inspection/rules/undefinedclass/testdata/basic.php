<?php
namespace App;

use App\Model\User;
use Lib\Missing\Thing;

$user = new <error descr="Fix the name or install the package declaring App\Model\Usr.">Model\Usr</error>();
$thing = new <error descr="Fix the name or install the package declaring Lib\Missing\Thing.">Thing</error>;
<error descr="Fix the name or install the package declaring App\Registry.">Registry</error>::boot();
echo <error descr="Fix the name or install the package declaring App\Config.">Config</error>::VERSION;
echo <error descr="Fix the name or install the package declaring App\Settings.">Settings</error>::$debug;
$date = new <error descr="Fix the name or install the package declaring DateTim.">\DateTim</error>();

class Admin extends <error descr="Fix the name or install the package declaring App\BaseUser.">BaseUser</error> implements <error descr="Fix the name or install the package declaring App\Auditable.">Auditable</error>, \Countable
{
    use <error descr="Fix the name or install the package declaring App\Timestamps.">Timestamps</error>;

    public function count(): int { return 0; }
}

interface Reader extends <error descr="Fix the name or install the package declaring App\Source.">Source</error> {}

$known = new User();
$known = new \ArrayObject();
echo User::class;

// A check whose result does not guard the code.
function unguarded()
{
    $loaded = extension_loaded('acme');
    return new <error descr="Fix the name or install the package declaring App\Unguarded.">Unguarded</error>($loaded);
}

// An earlier check about another value, name or extension.
function run($cb)
{
    if (!is_callable($cb)) {
        throw new \InvalidArgumentException();
    }
    if (!class_exists(\Vendor\Other\Thing::class) || !defined('OTHER_READY') || !extension_loaded('other')) {
        return null;
    }
    return new <error descr="Fix the name or install the package declaring App\Missingg.">Missingg</error>($cb);
}

// A condition about another extension.
if (extension_loaded('other')) {
    $other = new <error descr="Fix the name or install the package declaring Acme\Other\Handle.">\Acme\Other\Handle</error>();
}
if (class_exists($dynamic)) {
    $dynamic = new <error descr="Fix the name or install the package declaring Acme\Other\Dynamic.">\Acme\Other\Dynamic</error>();
}
