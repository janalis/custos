<?php
// Builtins whose stub parameter list differs from the names PHP declares.
namespace Drifted {
    session_set_cookie_params(lifetime_or_options: 3600, path: '/', domain: 'example.test', secure: true, httponly: true);
    function seek(\IntlCalendar $calendar): void { $calendar->set(year: 2026, month: 0, dayOfMonth: 1); }
    function row($dba) { return dba_fetch(key: 'k', dba: $dba); }
    function bind($ldap) { return ldap_sasl_bind(ldap: $ldap, dn: null, password: null, mech: 'GSSAPI'); }
    class Almanac extends \IntlCalendar {}
    function jump(Almanac $almanac): void { $almanac->set(year: 2026, month: 0, dayOfMonth: 1); }
}

// Extensions whose stub names are not checked against the extension itself.
namespace Unverified {
    snmp_set_quick_print(enable: true);
    function draw(\ImagickDraw $draw): void { $draw->setFontSize(point_size: 12.0); }
    function pixels(\Imagick $image) { return \ImagickPixelIterator::getPixelIterator(imagick: $image); }
}

// A subtype of the receiver renames the parameter: PHP binds the name
// against the method that actually runs.
namespace Overrides {
    class Renderer { public function render($template) { return "base:$template"; } }
    class TwigRenderer extends Renderer { public function render($view) { return "twig:$view"; } }
    function show(Renderer $r) { return $r->render(view: 'home'); }

    interface Handler { public function handle($request); }
    final class Impl implements Handler { public function handle($req) { return "impl:$req"; } }
    function run(Handler $h) { return $h->handle(req: 'x'); }

    abstract class Tpl { abstract protected function build($a); public function go() { return $this->build(b: 1); } }
    abstract class MidTpl extends Tpl {}
    final class TplImpl extends MidTpl { protected function build($b) { return "built:$b"; } }

    class Fac { public static function make($a) { return 'fac'; } public static function call() { return static::make(b: 1); } }
    class Fac2 extends Fac { public static function make($b) { return "fac2:$b"; } }

    // Diamond: the shared implementation is visited once.
    interface Left extends Handler {}
    interface Right extends Handler {}
    class Both implements Left, Right { public function handle($payload) { return $payload; } }
    function both(Handler $h) { return $h->handle(payload: 1); }

    // An override collecting named arguments accepts any name.
    class Sink { public function put($item) {} }
    class OpenSink extends Sink { public function put(...$items) {} }
    function fill(Sink $s) { $s->put(anything: 1); }

    // The override comes from a trait of the subclass.
    trait Loud { public function say($loudly) {} }
    class Voice { public function say($quietly) {} }
    class Shout extends Voice { use Loud; }
    function speak(Voice $v) { $v->say(loudly: 1); }
}

// A trait method runs in the using class, which may declare the callee.
namespace Traits {
    trait Builds {
        abstract protected function build($a);
        public function go() { return $this->build(b: 1) . self::make(d: 1) . static::make(d: 1); }
        public static function make($c) { return 'trait'; }
    }
    final class Built { use Builds; protected function build($b) { return "built:$b"; } public static function make($d) { return "own:$d"; } }
}

// Receivers that do not resolve to one known declaration.
namespace Uncertain {
    class Plain { public function a($x) {} }
    /** @method void virtual(int $x) */
    class Magic { public function __call($name, $arguments) {} }
    function magic(Magic $m) { $m->virtual(y: 1); }
    function missing(Plain $p) { $p->nothing(x: 1); }
    function ghost(Ghost $g) { $g->a(y: 1); }
    function untyped($o) { $o->a(y: 1); }
    function dynamic(Plain $p, $name) { $p->$name(y: 1); Plain::$name(y: 1); $name::a(y: 1); }
    class Other { public function a($y) {} }
    function either(Plain|Other $o) { $o->a(y: 1); }
    function half(Plain|Ghost $o) { $o->a(z: 1); }
}

// A function declared twice accepts the names of either declaration.
namespace Conditional {
    if (\PHP_VERSION_ID > 80000) { function cond($a) {} } else { function cond($b) {} }
    cond(b: 1);
    if (\PHP_VERSION_ID > 80000) { function open($a) {} } else { function open(...$all) {} }
    open(b: 1);
}
