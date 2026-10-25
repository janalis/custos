<?php
final class DotenvTest extends \PHPUnit\Framework\TestCase
{
    public function testLoad(): void
    {
        @mkdir($dir = sys_get_temp_dir() . '/dotenv');
        $path = tempnam($dir, 'env-');
        file_put_contents($path, 'FOO=BAR');
    }
}

final class EnvWriter
{
    public function write(string $dir, string $secret): void
    {
        $path = tempnam($dir, 'env-');
        <warning descr="Verify the temporary file remains in the required directory.">file_put_contents($path, $secret)</warning>;
    }
}
