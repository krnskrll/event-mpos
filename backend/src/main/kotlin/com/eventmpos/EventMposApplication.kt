package com.eventmpos

import org.springframework.boot.autoconfigure.SpringBootApplication
import org.springframework.boot.runApplication
import org.springframework.scheduling.annotation.EnableScheduling

@SpringBootApplication
@EnableScheduling
class EventMposApplication

fun main(args: Array<String>) {
    runApplication<EventMposApplication>(*args)
}
