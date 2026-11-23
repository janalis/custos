<?php
// Reports what this PHP binary deprecates, one name per line with its kind:
// f:function, m:Class::method, k:Class::CONSTANT, c:CONSTANT.
//
//   probe.php list    every deprecated name this binary declares, with the
//                     extension declaring it after a tab
//   probe.php check   of the names read from stdin, those it deprecates,
//                     then x:extension for each extension loaded

function deprecatedOnUse(string $name): bool
{
    $deprecated = false;
    set_error_handler(function (int $no) use (&$deprecated): bool {
        $deprecated = $deprecated || $no === E_DEPRECATED || $no === E_USER_DEPRECATED;
        return true;
    });
    try {
        constant($name);
    } catch (Throwable $e) {
    }
    restore_error_handler();
    return $deprecated;
}

function deprecated(string $entry): bool
{
    $name = substr($entry, 2);
    switch ($entry[0]) {
        case 'f':
            return function_exists($name) && (new ReflectionFunction($name))->isDeprecated();
        case 'm':
            [$class, $method] = explode('::', $name, 2);
            return method_exists($class, $method) && (new ReflectionMethod($class, $method))->isDeprecated();
        case 'k':
            [$class] = explode('::', $name, 2);
            return class_exists($class) && defined($name) && deprecatedOnUse($name);
        case 'c':
            return defined($name) && deprecatedOnUse($name);
    }
    return false;
}

$out = [];
if (($argv[1] ?? '') === 'list') {
    foreach (get_defined_functions()['internal'] as $f) {
        $rf = new ReflectionFunction($f);
        if ($rf->isDeprecated()) {
            $out[] = "f:$f\t" . strtolower((string) $rf->getExtensionName());
        }
    }
    foreach (get_declared_classes() as $c) {
        $rc = new ReflectionClass($c);
        if (!$rc->isInternal()) {
            continue;
        }
        foreach ($rc->getMethods() as $m) {
            if ($m->getDeclaringClass()->getName() === $c && $m->isDeprecated()) {
                $out[] = "m:$c::" . $m->getName() . "\t" . strtolower((string) $rc->getExtensionName());
            }
        }
        if (method_exists('ReflectionClassConstant', 'isDeprecated')) {
            foreach ($rc->getReflectionConstants() as $k) {
                if ($k->getDeclaringClass()->getName() === $c && $k->isDeprecated()) {
                    $out[] = "k:$c::" . $k->getName() . "\t" . strtolower((string) $rc->getExtensionName());
                }
            }
        }
    }
    if (class_exists('ReflectionConstant')) {
        foreach (get_defined_constants(true) as $ext => $constants) {
            foreach ($constants as $n => $v) {
                if ((new ReflectionConstant($n))->isDeprecated()) {
                    $out[] = "c:$n\t" . strtolower($ext);
                }
            }
        }
    }
} else {
    foreach (file('php://stdin', FILE_IGNORE_NEW_LINES | FILE_SKIP_EMPTY_LINES) as $entry) {
        if (deprecated($entry)) {
            $out[] = $entry;
        }
    }
    foreach (get_loaded_extensions() as $ext) {
        $out[] = 'x:' . strtolower($ext);
    }
}
sort($out);
echo implode("\n", $out), "\n";
