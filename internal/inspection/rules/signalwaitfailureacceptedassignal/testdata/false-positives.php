<?php
 $signal=pcntl_sigwaitinfo([SIGTERM],$info);if($signal>0){handleSignal($signal);}if(pcntl_sigwaitinfo([SIGTERM],$info)){}if($unknown){handleSignal($unknown);}function dead(){return;if($s=pcntl_sigwaitinfo([],$info)){handleSignal($s);}}

strlen("unrelated");

if($s=pcntl_sigwaitinfo([],$info)){$s=5;handler($s);}if($s=pcntl_sigwaitinfo([],$info)){if($unknown){handler($other);}}

function rejectedEarlier($set)
{
    $sig = pcntl_sigwaitinfo($set, $info);
    if ($sig === -1) {
        return;
    }
    if ($sig) {
        handle($sig);
    }
}

function rejectedInside($set)
{
    $sig = pcntl_sigwaitinfo($set);
    if ($sig) {
        if ($sig > 0) {
            handle($sig);
        }
    }
}

function rejectedInsideAssignedForm($set)
{
    if ($sig = pcntl_sigwaitinfo($set)) {
        if ($sig < 0) {
            throw new RuntimeException('wait failed');
        }
        handle($sig);
    }
}

function reboundBeforeCall($set, $box)
{
    if ($sig = pcntl_sigwaitinfo($set)) {
        $box->normalize($sig);
        handle($sig);
    }
}
