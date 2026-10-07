<?php
namespace Payload {
    use function Helpers\compact;

    function build($id, $tag) {
        return <weak_warning descr="Replace with '\compact('id', 'tag')'.">[</weak_warning>'id' => $id, 'tag' => $tag];
    }
}
