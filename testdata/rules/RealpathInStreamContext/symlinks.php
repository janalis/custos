<?php
// realpath() climbs from a symlink's target, dirname() from the link: only
// bases PHP already resolved (__DIR__, __FILE__, dirname() of them) are
// rewritten.
class CoreUpdate
{
    private string $symlinkToCoreFiles = '/var/www/public/typo3_src';

    public function target(string $version): string
    {
        return @<warning descr="Use 'dirname($this->symlinkToCoreFiles)' instead: realpath() fails inside stream wrappers.">realpath($this->symlinkToCoreFiles . '/../')</warning> . '/typo3_src-' . $version;
    }

    public function own(): string
    {
        return <warning descr="Use 'dirname(dirname(__FILE__, 2)) . '/x'' instead: realpath() fails inside stream wrappers.">realpath(dirname(__FILE__, 2) . '/../x')</warning>
            . <warning descr="Use 'dirname(dirname(...)) . '/y'' instead: realpath() fails inside stream wrappers.">realpath(dirname(...) . '/../y')</warning>
            . <warning descr="Use 'dirname(dirname(__LINE__)) . '/z'' instead: realpath() fails inside stream wrappers.">realpath(dirname(__LINE__) . '/../z')</warning>;
    }
}
