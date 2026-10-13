<?php
// The first value may be read before the second write: by a later statement
// of the 'if' block, or by a call when the target is shared state.

function cached_route(string $key, bool $warm)
{
    if ($warm) {
        $route = fetch_cached($key, $hit);
        if ($hit) {
            return $route;
        }
    }
    $route = build_route($key);
    return $route;
}

function page_script(bool $withMenu)
{
    if ($withMenu) {
        $script = 'menu();';
        queue_asset('menu');
        print_inline('menu', $script);
    }
    $script = 'tick();';
    print_inline('tick', $script);
}

function pick_label(array $shelf, string $key)
{
    if (isset($shelf[$key])) {
        $label = 'found';
        if (is_object($shelf[$key])) {
            return clone $shelf[$key];
        } else {
            return $shelf[$key];
        }
    }
    $label = 'absent';
    return $label;
}

class Reader
{
    private int $state = 0;

    public function read(bool $tail): void
    {
        if ($tail) {
            $this->state = 2;
            $this->consume();
        }
        $this->state = 1;
    }

    public function reset(): void
    {
        $this->state = 3;
        $this->state = $this->computeState();
    }

    private function consume(): void { echo $this->state; }

    private function computeState(): int { return $this->state + 1; }
}

function reload_table()
{
    global $table;
    $table = read_saved_table();
    $table = tidy_table();
    return $table;
}

function reload_cart()
{
    $_SESSION['cart'] = [];
    $_SESSION['cart'] = merge_session_cart();
}

function fill_out(array &$out)
{
    $out = [];
    $out = transform_out();
}

function fill_alias(array $rows)
{
    $alias = &$rows;
    $alias = [1];
    $alias = transform_alias();
    $later = function () { $alias = 0; return $alias; };
    return [$rows, $later];
}

function counter_value()
{
    static $hits;
    $hits = 0;
    $hits = next_hits();
    return $hits;
}

function dynamic_name(string $n)
{
    $$n = 1;
    $$n = compute_dynamic();
}

$fileScope = 1;
$fileScope = compute_file_scope();

function still_local()
{
    $local = 1;
    <error descr="$local is overwritten right after being assigned.">$local</error> = compute_local();
    return $local;
}

function shared_no_call()
{
    global $flag;
    $flag = 1;
    <error descr="$flag is overwritten right after being assigned.">$flag</error> = 2;
}

function not_read_later(bool $c)
{
    if ($c) {
        $mode = 'a';
        log_line('x');
    }
    <error descr="$mode is overwritten right after the 'if'; an 'else' may be missing.">$mode = 'b'</error>;
    return $mode;
}
