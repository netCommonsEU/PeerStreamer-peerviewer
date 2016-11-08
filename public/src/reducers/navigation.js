const SET_DRAWER_STATE = 'navigation/SET_DRAWER_STATE';
const TOGGLE_DRAWER_STATE = 'navigation/TOGGLE_DRAWER_STATE';

const navigation = (state = {drawerOpen: false}, action) => {
    switch (action.type) {
    case SET_DRAWER_STATE:
        return {...state, drawerOpen: action.state};
    case TOGGLE_DRAWER_STATE:
        if (state.drawerOpen) return {...state, drawerOpen: false};
        else return {...state, drawerOpen: true};
    default:
        return state;
    }
};

export default navigation;

export const setDrawerState = (state) => ({type: SET_DRAWER_STATE, state});
export const toggleDrawer = () => ({type: TOGGLE_DRAWER_STATE});