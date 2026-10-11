<?php
function breakOutOfDoWhile(array $found, array $candidates): array
{
    $found[] = 'seed';
    do {
        if ($candidates !== []) {
            break;
        }
        return [];
    } while (false);
    return $found;
}

function continueInDoWhile(array $found, $skip)
{
    $found[] = 1;
    if ($skip) {
        do {
            if ($skip > 1) {
                continue;
            }
            throw new \RuntimeException('none');
        } while (false);
    } else {
        return null;
    }
    return $found;
}

function gotoLabel(array $found): array
{
    $found[] = 'seed';
    goto done;
    done:
    return $found;
}

function doWhileWithoutEscape(array $found, $again)
{
    <weak_warning descr="Value is only written here and never read; the write is lost.">$found[]</weak_warning> = 1;
    do {
        $handler = function () {
            do {
                break;
            } while (true);
        };
        return $handler;
    } while ($again);
    return $found;
}

function afterReturn(array $found)
{
    <weak_warning descr="Value is only written here and never read; the write is lost.">$found[]</weak_warning> = 1;
    return 0;
    return $found;
}

function nestedScopesHaveTheirOwnLabels(array $found)
{
    <weak_warning descr="Value is only written here and never read; the write is lost.">$found[]</weak_warning> = 1;
    $jump = function () {
        goto end;
        end:
        return 1;
    };
    return $jump;
    return $found;
}

function caseCondition(array $found, $x)
{
    $found[] = 1;
    switch ($x) {
        case count($found):
            return 1;
    }
    return 0;
}
