<?php
namespace Payload {
    use function Helpers\compact;

    function build($id, $tag) {
        return \compact('id', 'tag');
    }
}
