package com.eventmpos.config

import com.eventmpos.auth.JwtService
import org.springframework.context.annotation.Configuration
import org.springframework.messaging.Message
import org.springframework.messaging.MessageChannel
import org.springframework.messaging.simp.config.ChannelRegistration
import org.springframework.messaging.simp.config.MessageBrokerRegistry
import org.springframework.messaging.simp.stomp.StompCommand
import org.springframework.messaging.simp.stomp.StompHeaderAccessor
import org.springframework.messaging.support.ChannelInterceptor
import org.springframework.messaging.support.MessageHeaderAccessor
import org.springframework.web.socket.config.annotation.EnableWebSocketMessageBroker
import org.springframework.web.socket.config.annotation.StompEndpointRegistry
import org.springframework.web.socket.config.annotation.WebSocketMessageBrokerConfigurer

/**
 * Real-time order push for the warehouse dashboard (EPIC-09). STOMP over a SockJS-capable endpoint
 * at /ws; the server broadcasts to /topic/events/{eventId}/orders.
 *
 * The STOMP CONNECT frame is authenticated: the client must send a valid JWT in the `Authorization`
 * header, or the connection is rejected. This closes the gap of an unauthenticated socket accepting
 * order data (the plain /ws handshake is permitted in SecurityConfig; auth happens here at CONNECT).
 */
@Configuration
@EnableWebSocketMessageBroker
class WebSocketConfig(
    private val jwtService: JwtService,
) : WebSocketMessageBrokerConfigurer {

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

    /** Validate the JWT carried in the STOMP CONNECT frame; reject the connection if it's invalid. */
    override fun configureClientInboundChannel(registration: ChannelRegistration) {
        registration.interceptors(object : ChannelInterceptor {
            override fun preSend(message: Message<*>, channel: MessageChannel): Message<*> {
                val accessor = MessageHeaderAccessor.getAccessor(message, StompHeaderAccessor::class.java)
                if (accessor != null && StompCommand.CONNECT == accessor.command) {
                    val header = accessor.getFirstNativeHeader("Authorization")
                    val token = header?.removePrefix("Bearer ")?.trim()
                    val claims = token?.let { jwtService.parse(it) }
                        ?: throw IllegalArgumentException("Unauthorized WebSocket connection: missing/invalid token")
                    // Attach a Principal so the session is associated with the user.
                    accessor.user = java.security.Principal { claims.subject }
                }
                return message
            }
        })
    }
}
