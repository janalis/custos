<?php
function owned(array $packages): array
{
    $owners = [];
    foreach ($packages as $package) {
        foreach ($package->classes as $class) {
            $owners[$class] = $package->id;
        }
        $owners = array_replace($owners, owned($package->children));
    }
    return $owners;
}
