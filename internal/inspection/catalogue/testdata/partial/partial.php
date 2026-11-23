<?php
declare(strict_types=1);

namespace App;

// Every line marked `// partial` builds a closure (PHP 8.6 partial function
// application) and calls nothing; the guard test compares the findings with
// the same lines written as first-class callables, `f(...)`.

final class Repo
{
    public function find(int $id, bool $lock = false): ?array
    {
        return $lock ? null : [$id];
    }

    public static function make(string $name, int $size): self
    {
        return new self();
    }

    public function __invoke(int $x): int
    {
        return $x;
    }

    /** @return list<\Closure> */
    public function own(): array
    {
        $c = [];
        $c[] = $this->find(?, true); // partial
        $c[] = self::make(?, 2); // partial
        $c[] = static::make('a', ...); // partial
        return $c;
    }
}

function load(string $path, int $flags = 0): string
{
    return $path . $flags;
}

function stop(string $message): never
{
    exit($message);
}

function gen(int $n): \Generator
{
    yield $n;
}

/**
 * @param resource $fp
 * @param list<string> $items
 * @return list<\Closure>
 */
function partials(Repo $repo, string $s, array $items, $fp, \PDO $pdo, \Fiber $fiber, $dir): array
{
    $c = [];
    $c[] = strlen(?); // partial
    $c[] = str_pad(?, 4); // partial
    $c[] = str_replace(?, ?, $s); // partial
    $c[] = in_array(?, $items); // partial
    $c[] = array_map(?, $items); // partial
    $c[] = array_search(?, $items); // partial
    $c[] = array_filter(?); // partial
    $c[] = json_decode(?, true); // partial
    $c[] = json_encode(?); // partial
    $c[] = preg_match(?, $s); // partial
    $c[] = preg_replace(?, '', $s); // partial
    $c[] = strpos(?, 'x'); // partial
    $c[] = substr(?, 0, 1); // partial
    $c[] = fseek($fp, ?); // partial
    $c[] = fread($fp, ?); // partial
    $c[] = fwrite($fp, ?); // partial
    $c[] = readdir(?); // partial
    $c[] = readdir($dir, ...); // partial
    $c[] = var_dump(?); // partial
    $c[] = print_r(?, true); // partial
    $c[] = exit(?); // partial
    $c[] = die(?); // partial
    $c[] = stop(?); // partial
    $c[] = load(?, 1); // partial
    $c[] = load(flags: 1, path: ?); // partial
    $c[] = gen(?); // partial
    $c[] = $repo->find(?, true); // partial
    $c[] = $repo->find(1, ...); // partial
    $c[] = Repo::make(?, 3); // partial
    $c[] = $repo(?); // partial
    $c[] = $pdo->prepare(?); // partial
    $c[] = $pdo->query(?); // partial
    $c[] = $fiber->start(?); // partial
    $c[] = $fiber->resume(?); // partial
    $c[] = sprintf(?, $s); // partial
    $c[] = implode(?, $items); // partial
    $c[] = is_null(?); // partial
    $c[] = intval(?, 16); // partial
    $c[] = assert(?); // partial
    $c[] = unserialize(?); // partial
    $c[] = mt_rand(?, 10); // partial
    $c[] = array_key_exists(?, $items); // partial
    $c[] = count(?); // partial
    $c[] = sizeof(?); // partial
    $c[] = clone(?); // partial
    $c[] = clone($repo, ?); // partial
    $c[] = call_user_func(?, 1); // partial
    $c[] = file_get_contents(?); // partial
    $c[] = file_put_contents(?, $s); // partial
    $c[] = ob_start(?); // partial
    $c[] = ftp_nb_get(?, 'a', 'b'); // partial
    $c[] = mysqli_query(?, 'SELECT 1'); // partial
    $c[] = unlink(?); // partial
    $c[] = array_merge(?, $items); // partial
    $c[] = strtolower(?); // partial
    $c[] = trim(?); // partial
    $c[] = trigger_error(?, E_USER_ERROR); // partial
    $c[] = header(?); // partial
    $c[] = session_start(?); // partial
    $c[] = date(?); // partial
    $c[] = sleep(?); // partial
    $c[] = rand(?, 10); // partial
    $c[] = array_walk($items, ?); // partial
    $c[] = usort($items, ?); // partial
    $c[] = extract(?); // partial
    $c[] = compact(?); // partial
    return $c;
}

final class ExampleTest extends \PHPUnit\Framework\TestCase
{
    public function testPartials(): void
    {
        $c = [];
        $c[] = $this->assertSame(?, 1); // partial
        $c[] = $this->assertTrue(?); // partial
        $c[] = self::assertNull(?); // partial
        $c[] = $this->assertCount(?, []); // partial
        $c[] = $this->expectException(?); // partial
        self::assertCount(5, $c);
    }
}
