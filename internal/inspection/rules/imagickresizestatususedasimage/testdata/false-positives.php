<?php
$i=new Imagick();$i->resizeImage(80,80,1,1);$i->getImageBlob();$other->getImageBlob();function dead(){return;$i=new Imagick();$i->resizeImage(1,1,1,1)->getImageBlob();}

strlen("unrelated");

function reboundByForeach(Imagick $image, array $list) {
    $result = $image->resizeImage(1, 1, Imagick::FILTER_LANCZOS, 1);
    foreach ($list as $result) {
        echo $result->count();
    }
}

function reboundByDestructuring(Imagick $image) {
    $result = $image->resizeImage(1, 1, Imagick::FILTER_LANCZOS, 1);
    [$result, $other] = pair();
    echo $result->count();
}

function reboundThroughAlias(Imagick $image) {
    $result = $image->resizeImage(1, 1, Imagick::FILTER_LANCZOS, 1);
    $alias = &$result;
    $alias = $image;
    echo $result->getNumberImages();
}
