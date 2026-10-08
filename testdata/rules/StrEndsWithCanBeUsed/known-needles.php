<?php
const ARCHIVE = '.zip';
class Paths
{
    const SUFFIX = '.tar';

    public function check(string $file, string $any): array
    {
        $ext = '.gz';
        $maybe = $any ? '.bz2' : '';
        return [
            <weak_warning descr="Replace with 'str_ends_with($file, $ext)'.">substr($file, -strlen($ext)) === $ext</weak_warning>,
            <weak_warning descr="Replace with 'str_ends_with($file, '.' . $any)'.">substr($file, -strlen('.' . $any)) === '.' . $any</weak_warning>,
            <weak_warning descr="Replace with 'str_ends_with($file, $any . '')'.">substr($file, -strlen($any . '')) === $any . ''</weak_warning>,
            <weak_warning descr="Replace with 'str_ends_with($file, self::SUFFIX)'.">substr($file, -strlen(self::SUFFIX)) === self::SUFFIX</weak_warning>,
            <weak_warning descr="Replace with 'str_ends_with($file, ARCHIVE)'.">substr($file, -strlen(ARCHIVE)) === ARCHIVE</weak_warning>,
            // May be empty: reported without a fix.
            <weak_warning descr="Replace with 'str_ends_with($file, $maybe)'.">substr($file, -strlen($maybe)) === $maybe</weak_warning>,
            <weak_warning descr="Replace with 'str_ends_with($file, $any)'.">substr($file, -strlen($any)) === $any</weak_warning>,
            <weak_warning descr="Replace with 'str_ends_with($file, strtolower($any))'.">substr($file, -strlen(strtolower($any))) === strtolower($any)</weak_warning>,
        ];
    }
}
