<?php
class Runner
{
    public function chains(array $rows, $qb, $ctx): void
    {
        foreach ($rows as $row) {
            $qb->where('a')->bind('k', $row);
            $qb->query()->run();
            $ctx->console()->tick();
        }
    }

    public function fluentArgument(array $rows, $qb): void
    {
        foreach ($rows as $row) {
            $qb->delete('t')->setParameter('id', $row[0]);
            $qb->getQuery()->execute();
        }
    }

    public function guarded(array $items, $wanted): void
    {
        foreach ($items as $item) {
            if ($item !== $wanted) {
                continue;
            }
            notify($wanted);
        }
        foreach ($items as $item) {
            if ($item === null) {
                throw new \LogicException('null');
            }
            echo $wanted;
        }
        foreach ($items as $item) {
            foreach ($item as $part) {
                if ($part) {
                    continue;
                }
            }
            <weak_warning descr="Statement does not depend on the loop; move it out.">notify($wanted);</weak_warning>
        }
        foreach ($items as $item) {
            foreach ($item as $part) {
                if ($part) {
                    continue 2;
                }
            }
            notify($wanted);
        }
        foreach ($items as $item) {
            switch ($item) {
                case 1:
                    break;
            }
            $f = function () { return 1; };
            <weak_warning descr="Statement does not depend on the loop; move it out.">notify($wanted);</weak_warning>
        }
    }

    public function resets(array $nodes, array $p): void
    {
        foreach ($nodes as $node) {
            $level = 1;
            if (isset($p['depth'])) {
                $level = $p['depth'] + 1;
            }
            $node->setLevel($level);
        }
    }
}
function earlyExit(array $items, $wanted): void
{
    foreach ($items as $item) {
        if ($item === null) {
            throw new \LogicException('null');
            echo 'unreachable';
        }
        notify($wanted);
    }
}
