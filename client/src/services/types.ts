export interface PublicUser {
  id: string;
  clerkUserId: string;
  username: string;
}

export interface MeResponse {
  authenticated: true;
  user: PublicUser | null;
  onboardingComplete: boolean;
}

export interface ProfileResponse {
  user: PublicUser;
}

export interface ApiError {
  message: string;
  authenticated?: boolean;
}

export interface RoomUser {
  id: string;
  username: string;
}

export interface RoomMemberDTO {
  user: RoomUser;
  joinedAt: string;
}

export interface RoomDTO {
  id: string;
  roomCode: string;
  host: RoomUser;
  members: RoomMemberDTO[];
  currentVideoId: string | null;
  currentTitle: string | null;
  currentThumbnailUrl: string | null;
  currentDuration: number | null;
  isPlaying: boolean;
  playbackPosition: number;
  playbackUpdatedAt: string | null;
}

export interface CreateRoomResponse {
  room: RoomDTO;
}

export interface JoinRoomResponse {
  room: RoomDTO;
}

export interface RoomResponse {
  room: RoomDTO;
}

export interface LeaveRoomResponse {
  left: true;
  roomDeleted: boolean;
}

export interface PresenceMember {
  userId: string;
  username: string;
  online: boolean;
}

export interface PresenceSnapshot {
  members: PresenceMember[];
  hostUserId: string;
}

export interface YouTubeSearchResult {
  videoId: string;
  title: string;
  channelTitle: string;
  thumbnailUrl: string;
  duration: string;
}

export interface YouTubeSearchResponse {
  results: YouTubeSearchResult[];
}

export interface YouTubeVideoResponse {
  result: YouTubeSearchResult;
}

export interface YouTubePlaylistResponse {
  results: YouTubeSearchResult[];
  playlistTitle?: string;
}

export interface PlaybackSnapshot {
  videoId: string | null;
  title: string | null;
  thumbnailUrl: string | null;
  duration: number | null;
  isPlaying: boolean;
  position: number;
  updatedAt: string | null;
  serverTime: string;
}

export type PlaybackAction = "play" | "pause" | "seek";

export interface ChatMessage {
  id: string;
  roomId: string;
  userId: string;
  username: string;
  content: string;
  createdAt: string;
}

export interface QueueItem {
  videoId: string;
  title: string;
  thumbnailUrl: string;
  duration: string;
}

export type QueueSnapshot = QueueItem[];

export type HostChangeReason = "manual" | "disconnect" | "left";

export interface HostUpdatePayload {
  hostUserId: string;
  hostUsername: string;
  previousHostUserId: string;
  previousHostUsername: string;
  reason: HostChangeReason;
}

export interface HostPendingPayload {
  pending: boolean;
  hostUserId: string;
  hostUsername: string;
  successorUsername?: string;
  deadline?: string;
}
