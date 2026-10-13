<?php
function register_versions(array $package, array $version, Loader $loader): array
{
    if ($version['feature']) {
        $package['version'] = $version['feature'];
        $loader->load($package); // reads the element through the array
    }
    $package['version'] = $version['pretty'];

    $package['count'] = 1;
    $package['count'] = count($package);

    $opts = [];
    if ($version['x']) {
        $opts['a']['b'] = 1;
        $loader->load($opts['b']);
    }
    <error descr="$opts['a']['b'] is overwritten right after the 'if'; an 'else' may be missing.">$opts['a']['b'] = 2</error>;

    return [$package, $opts];
}
