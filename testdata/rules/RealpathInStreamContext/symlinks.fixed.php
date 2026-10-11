<?php
// realpath() climbs from a symlink's target, dirname() from the link: only
// bases PHP already resolved (__DIR__, dirname() of it or of __FILE__) are
// rewritten.
class CoreUpdate
{
    private string $symlinkToCoreFiles = '/var/www/public/typo3_src';

    public function target(string $version): string
    {
        return @realpath($this->symlinkToCoreFiles . '/../') . '/typo3_src-' . $version;
    }

    public function own(): string
    {
        return dirname(dirname(__FILE__, 2)) . '/x'
            . realpath(dirname(...) . '/../y')
            . realpath(dirname(__LINE__) . '/../z');
    }
}
