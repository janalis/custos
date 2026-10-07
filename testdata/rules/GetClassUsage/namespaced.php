<?php
namespace Acme\Util {
    function get_class($value = null) { return 'x'; }
}

namespace Acme\Util {
    function local(?\Order $order) {
        return get_class($order);
    }
}

namespace Acme\App {
    function viaQualified(?\Order $order) {
        return \Acme\Util\get_class($order);
    }

    function viaGlobal(?\Order $order) {
        return <warning descr="get_class() rejects null on PHP 7.2+; guard the argument.">get_class($order)</warning>;
    }

    function viaRoot(?\Order $order) {
        return <warning descr="get_class() rejects null on PHP 7.2+; guard the argument.">\get_class($order)</warning>;
    }
}

namespace Acme\Other {
    use function Acme\Util\get_class;

    function imported(?\Order $order) {
        return get_class($order);
    }
}
