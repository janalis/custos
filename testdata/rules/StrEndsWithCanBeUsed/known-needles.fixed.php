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
            str_ends_with($file, $ext),
            str_ends_with($file, '.' . $any),
            substr($file, -strlen($any . '')) === $any . '',
            str_ends_with($file, self::SUFFIX),
            str_ends_with($file, ARCHIVE),
            // May be empty: reported without a fix.
            substr($file, -strlen($maybe)) === $maybe,
            substr($file, -strlen($any)) === $any,
            substr($file, -strlen(strtolower($any))) === strtolower($any),
        ];
    }
}
