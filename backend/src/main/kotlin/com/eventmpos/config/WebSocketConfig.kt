package com.eventmpos.config

import org.springframework.context.annotation.Configuration
import org.springframework.messaging.simp.config.MessageBrokerRegistry
import org.springframework.web.socket.config.annotation.EnableWebSocketMessageBroker
import org.springframework.web.socket.config.annotation.StompEndpointRegistry
import org.springframework.web.socket.config.annotation.WebSocketMessageBrokerConfigurer

/**
 * Real-time order push for the warehouse dashboard (EPIC-09). STOMP over a SockJS-capable endpoint
 * at /ws; the server broadcasts to /topic/events/{eventId}/orders. The dashboard subscribes and
 * updates live; it also has a REST polling fallback for when the socket drops.
 *
 * Note: the /ws handshake is permitted in SecurityConfig (warehouse trusted network). Authenticating
 * the STOMP CONNECT with the JWT is a sensible hardening follow-up.
 */
@Configuration
@EnableWebSocketMessageBroker
class WebSocketConfig : WebSocketMessageBrokerConfigurer {

    override fun configureMessageBroker(registry: MessageBrokerRegistry) {
        registry.enableSimpleBroker("/topic")
        registry.setApplicationDestinationPrefixes("/app")
    }

    override fun registerStompEndpoints(registry: StompEndpointRegistry) {
        // Browser dashboard uses SockJS at /ws.
        registry.addEndpoint("/ws")
            .setAllowedOriginPatterns("http://localhost:*", "http://127.0.0.1:*")
            .withSockJS()
        // Raw STOMP-over-WebSocket at /ws-native (native clients / integration tests).
        registry.addEndpoint("/ws-native")
            .setAllowedOriginPatterns("http://localhost:*", "http://127.0.0.1:*", "*")
    }
}
