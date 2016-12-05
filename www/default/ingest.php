<?php
    // Ingest - a very simple consumer of Postback requests, putting those
    // requests into a queue for consumption by other agents.
    define('MAX_QUEUE_SIZE', 100000);

    // Redis setup.
    $redis = new Redis();
    $redis->connect('redis', 6379);
    $redis->auth('SHI/hel7');

    // Get request data.
    $json = file_get_contents('php://input');
    $obj = json_decode($json);

    // Response setup.
    header('Content-Type: application/json');
    $response_json = array(
        'redis_is_connected' => $redis->ping(),
        'postback_queue_size' => $redis->lSize('data')
    );

    // Push the postback data to the queue.
    if ($redis->lSize('data') <= MAX_QUEUE_SIZE) {
        $redis->lpush('data', $json);
    } else {
        header('Retry-After: 10');
        http_response_code(503);
        $response_json['error'] = 'MAX_QUEUE_SIZE ' . MAX_QUEUE_SIZE . ' reached';
    }
    echo json_encode($response_json);

?>
